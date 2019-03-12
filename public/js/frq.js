$(document).ready(function(){
    $('#data').hide();
    $('#download').on('click', function(){ 
        dwnld();
    });
    $('#search').submit(function (e) {
        e.preventDefault();
        upload();
    });
});

function dwnld(){
    $.ajax({
        url: 'ajaxreq',
        type: 'post',
        dataType: 'html',
        data: {download: '1'},
        success: function(data) {
            printJSON(data);          
        },
        error: function(error){
            alert(error);
        }
    });
    $('#data').show();
}

function printJSON(data) {
    var json = $.parseJSON(data)
    if (!json.Empty) {
        $('#pick').empty();
        $.each(json.Data, function(index, name){
            $('#pick').append(
                $('<option></option>').val(index).html(name)
            );
        })
    }
}

function upload(){
    $.ajax({
        url: 'ajaxreq',
        type: 'post',
        dataType: 'html',
        data: {name: $('#pick').find(":selected").text()},
        success: function(data) {
            $('#message').text(data)
        },
        error: function(error) {
            alert(error);
        }
    });
}